FROM frolvlad/alpine-glibc

COPY target/goauth /usr/local/bin

CMD ["/usr/local/bin/goauth"]
