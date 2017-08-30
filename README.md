# GitlabBot
GitlabBot is a gitlab merge request munger, with useful features.

Usage:
```
Golang implementation of Gitlab Munger, with following features:
- Commenting
- Merging
- LGTM Counts
Using Gitlab database as well as webhook on the project.

Usage:
  gitlabbot [flags]

Flags:
  -H, --dbhost string        Postgres database Hostname/IPAddress (default "localhost")
  -n, --dbname string        Gitlab database name (default "gitlabhq_production")
  -p, --dbpassword string    Gitlab database password (default "Aa111111")
  -P, --dbport int           Postgres database port number (default 5432)
  -u, --dbuser string        Gitlab database username (default "gitlab")
  -b, --gitlabbase string    Gitlab base url (default "http://localhost:10080")
  -t, --gitlabtoken string   Gitlab user token for API access (default "K8F8SZEHyq4Dm9osdTT3")
  -h, --help                 help for gitlabbot
  -l, --lgtmtreashold int    Number of LGTMs required to merge the request (default 2)
  -v, --version              Prints version info

```

Setup: 
1. Create an access token with and elavated user on the project in gitlab.
2. Clone and build.
3. Run the binary specifying the required parameters.
4. Go to gitlab, Create a merge request, You should see the bot commenting help messages there.
