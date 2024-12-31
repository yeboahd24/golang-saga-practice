FROM postgres:13

COPY init-scripts/ /init-scripts/
COPY entrypoint.sh /entrypoint.sh

RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
