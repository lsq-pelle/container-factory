FROM progrium/busybox
ADD container-factory /
ADD packs /packs
ADD .dockercfg /root/
EXPOSE 3000
ENTRYPOINT ["/container-factory"]
