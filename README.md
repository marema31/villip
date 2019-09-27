# villip
proxy filtering string from the answer of the proxified application

Useful for proxifiing containerized web application that returns hardcoded URLs not modifiable from configuration

All configuration is down via environment variables :


Variable | Effects 
-------------|---------------------------
VILLIP_URL | Base URL of the application to be filtered
VILLIP_TYPES | Comma separated list of Content-Types of the answers that will be filtered
VILLIP_FROM / VILLIP_TO | String to be replaced / String that will replace
VILLIP_FROM_<i> / VILLIP_TO_<i> | Other couple of string to be replaced (with <i> from 1 to ...) 
VILLIP_DEBUG | If not empty print more logs to stdout 

