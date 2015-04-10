FROM golang:onbuild
ADD packs /packs
ADD .dockercfg /root/
EXPOSE 3000
