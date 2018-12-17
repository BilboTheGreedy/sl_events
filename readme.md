# SL_EVENTS
SL Events is a collector of softlayer events and slack notification utility.

## How it works
Collector > DB > Announcer 

### Collector

The collector periodically polls softlayer api for new events and adds them to the database
Collector requires secrets from vault
Secret should contain: All Softlayer API accounts - Collector will loop through them all.

### Announcer

The announcer listens for notify events from database and posts to slack
Announcer requires secrets from vault
Secret should contain: slack_token slack_channel

### DB

The database stores new events and notify the announcer.


# How to install

Build the announcer and collector then run docker-compose build