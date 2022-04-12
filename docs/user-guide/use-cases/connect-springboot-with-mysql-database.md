# Connect SpringBoot with Mysql Database

## Introduction

This document will help you to deploy a sample Spring Boot Application, using **mysql Helm Chart**

### **1. Deploy a mysql Helm Chart**

To deploy mysql Helm Chart, you can refer to our documentation on [Deploy mysql Helm Chart](../deploy-chart/examples/deploying-mysql-helm-chart.md)

### **2. Fork the Git Repository**

For this example, we are using the following [GitHub Repo](https://github.com/devtron-labs/springboot), you can clone this repository and make following changes in the files.

#### _\*Configure application.properties_

Set the database configuration in this file.

```java
spring.datasource.url=jdbc:mysql://<service-name>/<mysql database-name>
spring.jpa.hibernate.ddl-auto=update
spring.jpa.show-sql=true
spring.datasource.username=<mysql-user>
spring.datasource.password=<mysql-password>
spring.datasource.driver-class-name=com.mysql.jdbc.Driver
spring.jpa.properties.hibernate.dialect=org.hibernate.dialect.MySQL5Dialect
spring.jpa.open-in-view=true
```

#### _Configure the Dockerfile_

```bash
# syntax=docker/dockerfile:experimental
FROM maven:3.5-jdk-8-alpine as build
WORKDIR /workspace/app

COPY pom.xml .

RUN mvn -B -e -C -T 1C org.apache.maven.plugins:maven-dependency-plugin:3.0.2:go-offline

COPY . .
RUN mvn clean package -Dmaven.test.skip=true


FROM openjdk:8-jdk-alpine
RUN addgroup -S demo && adduser -S demo -G demo
VOLUME /tmp
USER demo
ARG DEPENDENCY=/workspace/app/target/dependency
COPY --from=build /workspace/app/target/docker-demo-0.0.1-SNAPSHOT.jar app.jar
ENTRYPOINT ["java","-jar", "app.jar"]
```

### **3. Create Application on Devtron**

To learn how to create an application on Devtron, refer to our documentation on [Creating Application](../creating-application/)

#### _\*Git Repository_

In this example, we are using the url of the forked Git repository.

#### _\*Docker configuration_

Give, the path of the Dockerfile.

#### _\*\*\_Configure Deployment Template_\*\_

Enable `Ingress`, and give the path on which you want to host the application.

![](../../.gitbook/assets/three%20%282%29.jpg)

#### _\*Set up the CI/CD Pipelines_

Set up the CI/CD pipelines. You can set them to trigger automatically or manually.

#### _\*Trigger Pipelines_

Trigger the CI Pipeline, build should be **Successful**. Then trigger the CD Pipeline, deployment pipeline will be initiated, after some time the status should be **Healthy**.

### **4. Final Step**

#### _\*Test Rest API_

It exposes 3 REST endpoints for it's users to create, to _view specific_ student record and _view all_ student records.

To test Rest API, you can use _curl_ command line tool

_**Create a new Student Record**_

Create a new POST request to create a new Transaction. Once the transaction is successfully created, you will get the _student id_ as a response.

Curl Request is as follows:

```bash
sudo curl -d '{"name": "Anushka", "marks": 98}' -H "Content-Type: application/json" -X POST http://<hostname>/<path-name>/create
```

_**View All Student's Data**_

To view all student records, GET Request is:

_**path**_ will be the one that you have given in Step 3 while configuring the Deployment Template.

`http://<hostname>/<path>/viewAll`

![](../../.gitbook/assets/use-cases-springboot-view-student-data.jpg)

_**View student's data By student ID**_

To view student data by student id, GET Request is:

`http://<hostname>/<path>/view/<id>`

_**path**_ will be the one that you have given in Step 3 while configuring the Deployment Template.

![](../../.gitbook/assets/use-cases-springboot-view-student-data-with-id%20%282%29.jpg)

