FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-mongodb"]
COPY baton-mongodb /