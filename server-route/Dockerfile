FROM alpine
ARG module_name
ENV MODULE_NAME ${module_name}
ADD VERSION .
ADD target/$module_name /
CMD ./$MODULE_NAME