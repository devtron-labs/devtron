ALTER TABLE ci_workflow
    ADD column IF NOT EXISTS executor_type varchar(50);