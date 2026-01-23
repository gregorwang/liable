param(
  [string]$BaseUrl,
  [string]$ApiKey,
  [string]$Model,
  [string]$Comment = "这是一条测试评论"
)

$dotenv = @{}
if (Test-Path ".env") {
  Get-Content ".env" | ForEach-Object {
    $line = $_.Trim()
    if ($line -eq "" -or $line.StartsWith("#")) {
      return
    }
    if ($line -match "^\s*([^=]+)=(.*)$") {
      $key = $matches[1].Trim()
      $value = $matches[2].Trim()
      if ($value.StartsWith('"') -and $value.EndsWith('"')) {
        $value = $value.Substring(1, $value.Length - 2)
      }
      $dotenv[$key] = $value
    }
  }
}

function Get-ConfigValue([string]$key) {
  $envValue = [System.Environment]::GetEnvironmentVariable($key)
  if ($envValue) {
    return $envValue
  }
  if ($dotenv.ContainsKey($key)) {
    return $dotenv[$key]
  }
  return $null
}

if (-not $BaseUrl) {
  $BaseUrl = Get-ConfigValue "AI_BASE_URL"
  if (-not $BaseUrl) {
    $BaseUrl = Get-ConfigValue "OPENAI_BASE_URL"
  }
}

if (-not $ApiKey) {
  $ApiKey = Get-ConfigValue "AI_API_KEY"
  if (-not $ApiKey) {
    $ApiKey = Get-ConfigValue "OPENAI_API_KEY"
  }
}

if (-not $Model) {
  $Model = Get-ConfigValue "AI_MODEL"
  if (-not $Model) {
    $Model = Get-ConfigValue "OPENAI_MODEL"
  }
}

if (-not $BaseUrl -or -not $ApiKey -or -not $Model) {
  Write-Error "Missing config: base_url, api_key, or model. Provide params or set AI_* / OPENAI_*."
  exit 1
}

if ($BaseUrl -notmatch "/v1/?$") {
  Write-Warning "Base URL does not end with /v1. The client appends /chat/completions."
}

$endpoint = $BaseUrl.TrimEnd("/") + "/chat/completions"
Write-Host "Testing AI endpoint: $endpoint"
Write-Host "Model: $Model"

$payload = @{
  model = $Model
  messages = @(
    @{
      role = "system"
      content = "You are an assistant that reviews user comments for policy compliance. Return JSON: is_approved, tags, reason, confidence."
    },
    @{
      role = "user"
      content = "Comment:`n$Comment"
    }
  )
  temperature = 0.2
  response_format = @{
    type = "json_object"
  }
} | ConvertTo-Json -Depth 6

try {
  $response = Invoke-RestMethod -Method Post -Uri $endpoint -Headers @{
    "Authorization" = "Bearer $ApiKey"
    "Content-Type" = "application/json"
  } -Body $payload -TimeoutSec 30
} catch {
  Write-Error "Request failed: $($_.Exception.Message)"
  if ($_.Exception.Response) {
    $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
    $body = $reader.ReadToEnd()
    if ($body) {
      Write-Host "Response body:"
      Write-Host $body
    }
  }
  exit 1
}

if (-not $response.choices -or $response.choices.Count -eq 0) {
  Write-Error "Response missing choices."
  exit 1
}

$content = $response.choices[0].message.content
Write-Host "AI response content:"
Write-Host $content
