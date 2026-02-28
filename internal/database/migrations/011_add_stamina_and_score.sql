-- Migration: Add stamina and score tracking to guest_set_progress
-- Tracks player stamina for submitting answers and score for puzzle sets

ALTER TABLE guest_set_progress 
ADD COLUMN current_stamina INT DEFAULT 35,
ADD COLUMN last_stamina_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
ADD COLUMN current_score INT DEFAULT 500;

-- Add comments
COMMENT ON COLUMN guest_set_progress.current_stamina IS 'Current stamina level (used for submitting answers)';
COMMENT ON COLUMN guest_set_progress.last_stamina_update IS 'Last time stamina was updated/deducted';
COMMENT ON COLUMN guest_set_progress.current_score IS 'Current score for this puzzle set';