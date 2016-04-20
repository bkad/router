FROM alpine:edge

# install common packages
RUN apk add --update-cache \
  bash \
  caddy \
  curl \
  wget \
  && rm -rf /var/cache/apk/* \
  && touch /var/run/caddy.pid

# add router user
#RUN addgroup -S router && \
#  adduser -S -G router -H -h /opt/router -D router

COPY . /

# Fix some permission since we'll be running as a non-root user.
#RUN chown -R router:router /opt/router \
#  && chown router:router /var/run/caddy.pid \

# Some day figure out how to run this as non root.
#USER router

CMD ["/opt/router/sbin/boot"]
EXPOSE 80 443 2015 2222 9090

ENV DEIS_RELEASE 2.0.0
WORKDIR /opt/router
