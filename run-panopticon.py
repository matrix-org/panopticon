#!/usr/bin/env python
import os
from scripts.aggregate import Config

c = Config()
db = "%s:%s@tcp/%s" % (c.db_user, c.db_password, c.db_name)
command = "./panopticon --db-driver=mysql --db %s --port 34124" % db
os.system(command)
