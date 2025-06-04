ALTER TABLE deployment_template_history
    ADD COLUMN IF NOT EXISTS pipeline_ids integer[];
