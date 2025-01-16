#!/bin/bash

echo "Running pg-cron initialization" > /tmp/pg-cron-init.log

dbname="$POSTGRES_DB"
customconf=/var/lib/postgresql/data/custom-conf.conf
echo "" > $customconf
echo "shared_preload_libraries = 'pg_cron'" >> $customconf
echo "cron.database_name = '$dbname'" >> $customconf
chown postgres $customconf
chgrp postgres $customconf

# include custom config from main config
conf=/var/lib/postgresql/data/postgresql.conf
found=$(grep "include = '$customconf'" $conf)
if [ -z "$found" ]; then
  echo "include = '$customconf'" >> $conf
fi

echo "Completed pg-cron initialization" >> /tmp/pg-cron-init.log
