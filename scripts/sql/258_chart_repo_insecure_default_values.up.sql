-- till this migration script, allow_insecure_connection is FALSE in Database
-- insecureSkipTlsVerification was not derived from Database and was hardcoded as True in Kubelink.
-- Now we are deriving it's value from DB and to preserve existing behaviour, migration values in DB are set to TRUE
update chart_repo set allow_insecure_connection=true;
