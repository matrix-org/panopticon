#!/usr/bin/env python
import os
from scripts.aggregate import Config

c = Config()
db = "%s:%s@tcp/%s" % (c.DB_USER, c.DB_PASSWORD, c.DB_NAME)
command = "./panopticon --db-driver=mysql --db %s --port 34124" % db
os.system(command)
