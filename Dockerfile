FROM alpine:3.10 as final
ADD build/faucet-service .
EXPOSE 8080
CMD ["./faucet-service"]
