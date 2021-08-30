FROM thiht/smocker

RUN apk add --no-cache curl
COPY start_smocker.sh /opt
RUN chmod a+x /opt/start_smocker.sh
WORKDIR /opt

CMD ["/bin/sh", "-c", "/opt/start_smocker.sh"]