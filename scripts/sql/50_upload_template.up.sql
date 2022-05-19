ALTER TABLE chart_ref
    ADD COLUMN chart_description varchar(250) DEFAULT '',
    ADD COLUMN user_uploaded boolean DEFAULT false ;