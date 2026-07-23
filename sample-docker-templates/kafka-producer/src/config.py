import os


# Database Info
db_host = os.getenv("db_host")
db_port = os.getenv("db_port")
db_user = os.getenv("db_user")
db_pass = os.getenv("db_pass")

# Common Info
logger_level = os.getenv("logger_level")

# JSON print related
json_sort_keys = True
json_indent = 4

