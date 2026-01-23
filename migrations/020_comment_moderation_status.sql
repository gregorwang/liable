-- Add moderation status tracking to comments
ALTER TABLE comment
ADD COLUMN IF NOT EXISTS moderation_status VARCHAR(30) NOT NULL DEFAULT 'pending';

DO $$
BEGIN
	IF NOT EXISTS (
		SELECT 1
		FROM information_schema.table_constraints
		WHERE table_name = 'comment'
		  AND constraint_name = 'comment_moderation_status_check'
	) THEN
		ALTER TABLE comment
		ADD CONSTRAINT comment_moderation_status_check
		CHECK (moderation_status IN ('pending', 'approved', 'rejected', 'pending_second_review'));
	END IF;
END $$;
