FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-mongodb-atlas"]
COPY baton-mongodb-atlas /