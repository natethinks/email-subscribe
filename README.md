# email-subscribe
Simple email subscriber store

This project is to generate a simple nosql store for email addresses to be submitted to. Email addresses are confirmed using regex as well as
SMTP requests (to SMTP servers that support it) to identify if the user exists. They are then stored in a nosql (boltdb) database.

#To do
- use simple config file for content and settings
- generate html template with the basic landing page and form for email submission.

