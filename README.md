# villip

#### Simple HTTP proxy that can do :
####  - string replacement on the content of proxyfied requests/responses
####  - header set/replacement of proxyfied requests/responses

Villip can be used for containerization of legacy web application that provide absolute path in the page by replacing all the absolute link to the correct one without modifying the application.

The replacement will be also done in Location header if the proxyfied site returns an HTTP 301 or 302 code.

Villip can also be used to replace or set a Header value in the HTTP request/reponse

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
VILLIP_INSECURE   | no        | If present Villip will not verify the tls certificate validity for proxified site
VILLIP_TO         | yes       | Replacement for the VILLIP_FROM string
VILLIP_FOR_XX     | no        | Comma separated list of urls concerned by this XX search
VILLIP_FROM_XX    | no        | XX string to search (XX = number starting at 1)
VILLIP_TO_XX      | no        | Replacement for the corresponding VILLIP_FROM_XX string
VILLIP_PORT       | no        | Port of proxy (8080 by default)
VILLIP_PRIORITY   | no        | Priority of the filter (0 by default)
VILLIP_RESTRICTED | no        | Comma separated list of networks authorized to use this proxy (no restriction if empty), localhost is always authorized
VILLIP_TYPES      | no        | Comma separated list of content type that will be filtered (by default text/html, text/css, application/javascript)
VILLIP_URL        | yes       | Base url of the proxyfied site

## YAML/JSON configuration files
Each YAML/JSON files in the folder pointed by VILLIP_FOLDER environment variable contains the configuration of a filter, the format of these files correspond the same parameter in environment variable formet.

The example below give the overall YAML structure of configuration file when using all attributes.

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
response:         #the http response config part
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
  header:
    - name: "X-community" # Beware that Vilip will Canonicalize your header name (Mime convention) X-ENV will be converted to X-Env
      value: "In real life"
      force: false  #if force = false the header value is set or replaced only if the header does not exist or if value is empty
      add: false # if true a new header line is added
request:        #the http request config part
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
  header:
    - name: "X-community"
      value: "In real life"
      force: false
restricted:
  - "192.168.1.0/24"
  - "192.168.8.0/24"
token:
  - header: X-MY-TOKEN
    value: "123"
    action: "accept"
  - header: X-MMYTOKEN
    value: "456"
    action: "accept"
  - header: X-MMYTOKEN
    value: "789"
    action: "reject"
  - header: X-MMYSECONDTOKEN
    value: "ABC"
    action: "accept"
  - header: X-MMYTHIRDTOKEN
    action: "notempty"
content-types:
  - "text/html"
  - "application/json"
```
## Conditionnal proxy
More than one file can refer to the same port, in this case all except one must have at a `token` or `restricted` attribute.
Villip will proxifies the request to one of the definition that will be fulfilled by the request condition (on header and/or source IP).
For `token` attribute, the condition on same header will be combined by logical `OR` but condition on different header are combined by logical `AND` operation.

# Disclaimer
I use this application for development environment, security was not a concern for this tool. Do not use it for production environment without being sure of what you are doing


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
