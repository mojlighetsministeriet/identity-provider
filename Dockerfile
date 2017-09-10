FROM scratch
COPY identity-provider /
ENV RSA_PRIVATE_KEY ""
ENV DATABASE_TYPE "mysql"
ENV DATABASE "user:password@/dbname?charset=utf8mb4,utf8&parseTime=True&loc=Local"
EXPOSE 1323
ENTRYPOINT ["/identity-provider"]
