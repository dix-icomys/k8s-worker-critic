FROM scratch

COPY build/critic /usr/local/bin/critic

CMD ["critic"]
