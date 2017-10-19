#!/bin/bash

# This is a dirty freaking hack. In order to get the mongo image to create the appropriate database
# from the environment we have to do this nasty --eval thing. Come on folks, support environment
# variable expansion!
mongo -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD admin --eval "dbname='$MONGO_DB', username='$MONGO_USER', dbpass='$MONGO_PASS'" /docker-entrypoint-initdb.d/mongo-init