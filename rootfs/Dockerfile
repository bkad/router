FROM alpine:3.3

# install common packages
RUN apk add --update-cache \
	bash \
	curl \
	geoip \
	libssl1.0 \
	openssl \
	pcre \
	sudo \
	&& rm -rf /var/cache/apk/*

# add router user
RUN addgroup -S router && \
	adduser -S -G router -H -h /opt/router -D router

COPY . /

# Fix some permission since we'll be running as a non-root user.
RUN chown -R router:router /opt/router \
  && touch /var/run/caddy.pid \
  && chown router:router /var/run/caddy.pid

USER router

CMD ["/opt/router/sbin/boot"]
EXPOSE 80 443 2015 2222 9090

ENV DEIS_RELEASE 2.0.0
WORKDIR /opt/router
