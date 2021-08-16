FROM alpine 

COPY ./contracts-discord /usr/bin//contracts-discord

CMD ["contracts-discord"]
