# GCP Golang Utilities

## Overview
GO Wrappers for some of the GCP API's we use
* pgdb
  * CloudSQL Postgresql
  * built around github.com/jackc/pgx
* email 
  * uses SendGrid Email service
  * built around github.com/sendgrid/sendgrid-go
* secrets 
  * uses GCP Secrets Manager
  * built around cloud.google.com/go/secretmanager/apiv1
* storage 
  * uses GCP Storage
  * built around cloud.google.com/go/storage
* util
  * general purpose routines (not specific to GCP)  

## Unit tests
The unit tests expect to obtain this information from a 'secrets' plain text file. The path to this file needs to
be specified in an environment variable with the key "utilities_config"
The format of the secrets data is a JSON structure and the unit tests require the following entries to execute successfully

````
{
  "ConfigName": "utilities",
  "Records": [
    {
      "Key": "CLOUDSQL_LOCAL",
      "Value": "user=xxx password=xxxx host=xx.xx.xx.xx port=xxxx dbname=xxxx sslmode=require"
    },
    {
      "Key": "TLS_CLIENT_KEY",
      "Value": "~\xxxkey.pem"
    },
    {
      "Key": "TLS_CLIENT_CERT",
      "Value": "~\xxxcert.pem"
    },
    {
      "Key": "FROM_USER",
      "Value": ""
    },
    {
      "Key": "SENDGRID_API_KEY",
      "Value": "xxxxxxx"
    },
    {
      "Key": "FROM_EMAIL",
      "Value": "no-reply@xxxxx.com"
    },
    {
      "Key": "EMAIL_RECIPIENT",
      "Value": "xxxxx@gmail.com"
    },
    {
      "Key": "CLOUD_STORAGE_BUCKET",
      "Value": "xxxxxx.appspot.com"
    },
    {
      "Key": "STORAGE_CREDENTIALS",
      "Value": "~\xxxxx_cred.json"
    }
  ]
}

````

## Using the secrets utility
email, pgdb, storage & util are not dependent on the secrets utility, except for the execution of unit tests (see above).

To use the secrets utility, you may 
  * Call InitializeFromEnvironment, passing in the name of an environment variable that will resolve to the path of the secrets file or
  * Call Initialize, passing in the path of the secrets file.
The path may refer to a local file or to a GCP Secrets identifier (to which you must already have access). 
 

