# villip

#### Simple HTTP proxy that can do string replacement on the content of proxyfied responses.

Villip can be used for containerization of legacy web application that provide absolute path in the page by replacing all the absolute link to the correct one without modifying the application.

The replacement will be also done in Location header if the proxyfied site returns an HTTP 301 or 302 code. 

# Usage
Configuration of Villip is done via environments variables or a folder containing YAML files. More than one filter can be described (one by environment variable and the others by configuration files). Each filter must be listening to a different TCP port. 

## Environment variables

Variable          | Mandatory |  Definition
------------------|-----------|---------------------
VILLIP_DEBUG      | no        | If present Villip will print debug logs
VILLIP_DUMPFOLDER | no        | If present Villip will dump the response (original and filtered) to files (two by requests)
VILLIP_DUMPURLS   | no        | If present Villip will dump the response (original and filtered) only for URLs correponding to one of the provided regular expression (commas-separated list), if DUMPFOLDER not provided the dump will be on STDOUT
VILLIP_FOLDER     | no        | Path to folder containing YAML configuration files, if present the other environment variables are no more mandatory
VILLIP_FOR        | no        | Comma separated list of urls concerned by the first search/replace (all if empty)
VILLIP_FROM       | yes       | First string to search
VILLIP_FORCE      | no        | If present Villip will ignore the content-type and filter all responses
VILLIP_TO         | yes       | Replacement for the VILLIP_FROM string
VILLIP_FOR_XX     | no        | Comma separated list of urls concerned by this XX search
VILLIP_FROM_XX    | no        | XX string to search (XX = number starting at 1)
VILLIP_TO_XX      | no        | Replacement for the corresponding VILLIP_FROM_XX string
VILLIP_PORT       | no        | Port of proxy (8080 by default)
VILLIP_RESTRICTED | no        | Comma separated list of networks authorized to use this proxy (no restriction if empty), localhost is always authorized
VILLIP_TYPES      | no        | Comma separated list of content type that will be filtered (by default text/html, text/css, application/javascript)
VILLIP_URL        | yes       | Base url of the proxyfied site

## YAML/JSON configuration files
Each YAML/JSON files in the folder pointed by VILLIP_FOLDER environment variable contains the configuration of a filter, the format of these files correspond the same parameter in environment variable formet.

```yaml
---
port: 8081
force: true
url: "http://localhost:1234/url1"
dump:
  folder: /var/log/villip/dump
  urls:
    - /books/
    - /movies/
replace:
  - from: "book"
    to: "smartphone"
    urls:
      - /youngster/
  - from: "dance"
    to: "chat"
    urls:
      - /youngsters/
      - /geeks/
  - from: "meeting"
    to: "texting"
restricted: 
  - "192.168.1.0/24"
  - "192.168.8.0/24"
content-types:
  - "text/html"
  - "application/json"
```

# Disclaimer
I use this application for development environment, security was not a concern for this tool. Do not use it for production environment without being sure of what you do


## Contribution
I've made this project as a real use case to learn Golang.
I've tried to adopt the Go mindset but I'm sure that other gophers could do better. 

If this project can be useful for you, feel free to open issues, propose documentation updates or even pull request.

Contributions are of course always welcome!

1. Fork marema31/villip (https://github.com/marema31/villip/fork)
2. Create a feature branch
3. Commit your changes
4. Create a Pull Request

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for details.

## License

Copyright (c) 2019-present [Marc Carmier](https://github.com/marema31)

Licensed under [BSD-2 Clause License](./LICENSE)