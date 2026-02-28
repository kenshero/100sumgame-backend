-- Migration: Add game_settings table
-- Stores configurable game parameters that can be updated without server restart

CREATE TABLE IF NOT EXISTS game_settings (
    key VARCHAR(100) PRIMARY KEY,
    value VARCHAR(255) NOT NULL,
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Insert default game settings
INSERT INTO game_settings (key, value, description) VALUES
('stamina_max', '35', 'Maximum stamina a player can have'),
('stamina_regen_interval_minutes', '5', 'Minutes to regenerate 1 stamina'),
('stamina_regen_amount', '1', 'Amount of stamina to regenerate per interval'),
('initial_score', '500', 'Initial score for a new puzzle set'),
('score_deduction_per_mistake', '10', 'Score deducted per wrong answer'),
('score_minimum', '0', 'Minimum score allowed (cannot go below this)');

-- Add comments
COMMENT ON TABLE game_settings IS 'Configurable game parameters';
COMMENT ON COLUMN game_settings.key IS 'Setting identifier (primary key)';
COMMENT ON COLUMN game_settings.value IS 'Setting value (stored as string)';
COMMENT ON COLUMN game_settings.description IS 'Description of what this setting does';