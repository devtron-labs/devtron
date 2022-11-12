-- Table Definition
CREATE TABLE "public"."chart_ref_metadata" (
    "chart_name" varchar(100) NOT NULL,
    "chart_description" text NOT NULL,
    PRIMARY KEY ("chart_name")
);

---Inserting Records-----
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Rollout Deployment', 'This chart deploys an advanced version of deployment that supports Blue/Green and Canary deployments. For functioning, it requires a rollout controller to run inside the cluster.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('CronJob & Job', 'This chart deploys Job & CronJob.  A Job is a controller object that represents a finite task and CronJob is used to schedule creation of Jobs.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Knative', 'This chart deploys Knative which is an Open-Source Enterprise-level solution to deploy Serverless apps.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Deployment', 'Creates a deployment that runs multiple replicas of your application and automatically replaces any instances that fail or become unresponsive.');
