# pg-sync

The purpose of pg-sync is to syncronize two or more tables
between two PostgreSQL databases, by truncating the destination
table and copying the source table as-is.

---

### Usage

pg-sync \
--config config.yaml \
--logfile /var/log/pg-sync.log \
--v

--config: the full path of the configuration file \
--logfile: the full path of the log file \
--v: write a verbose log (default is information)