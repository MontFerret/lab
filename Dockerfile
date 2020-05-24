# Build the final container. And install
FROM microbox/chromium-headless:83.0.4103.0 as runner

RUN apt-get update && apt-get install -y dumb-init

WORKDIR /root

# Add worker binary
COPY lab /bin/lab

EXPOSE 8080

ENTRYPOINT ["dumb-init", "--"]

CMD ["/bin/sh", "-c", "chromium --no-sandbox --disable-setuid-sandbox --disable-gpu --headless --remote-debugging-port=9222 & /bin/lab --wait http://127.0.0.1:9222/json/version"]
