-- Table Definition
CREATE TABLE "public"."chart_ref_metadata" (
    "chart_name" varchar(100) NOT NULL,
    "chart_description" text NOT NULL,
    PRIMARY KEY ("chart_name")
);

---Inserting Records-----
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Rollout Deployment', 'Chart to deploy an advanced version of Deployment that supports blue-green and canary deployments. It requires a rollout controller to run inside the cluster to function.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('CronJob & Job', 'Chart to deploy a Job/CronJob. Job is a controller object that represents a finite task and CronJob can be used to schedule creation of Jobs.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Knative', 'Chart to deploy an Open-Source Enterprise-level solution to deploy Serverless apps.');
INSERT INTO "chart_ref_metadata" ("chart_name", "chart_description") VALUES
    ('Deployment', 'Chart to deploy a Deployment that runs multiple replicas of your application and automatically replaces any instances that fail or become unresponsive.');