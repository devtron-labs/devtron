ALTER TABLE chart_ref
    ADD COLUMN chart_description text DEFAULT '',
    ADD COLUMN user_uploaded boolean DEFAULT false ;