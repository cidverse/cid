############################################################
# Base Image
############################################################

# Base Image
FROM amazoncorretto:17-alpine

##############################################################
## Environment
#############################################################

ENV JVM_USER_LANGUAGE="en" \
    JVM_USER_COUNTRY="US" \
    JVM_USER_TIMEZONE="UTC" \
    JVM_FILE_ENCODING="UTF8"

############################################################
# Installation
############################################################

# Copy files from rootfs to the container (there should only be one in /dist)
# TODO: replace with build arg for artifact
ADD dist/*.jar /app.jar

############################################################
# Execution
############################################################

# Expose
EXPOSE 8080/tcp

# Execution
CMD "java" \
    "-Djava.security.egd=file:/dev/./urandom" \
    "-Djava.net.useSystemProxies=true" \
    "-Duser.language=${JVM_USER_LANGUAGE}" \
    "-Duser.country=${JVM_USER_COUNTRY}" \
    "-Duser.timezone=${JVM_USER_TIMEZONE}" \
    "-Dfile.encoding=${JVM_FILE_ENCODING}" \
    "-XX:-TieredCompilation" \
    "-XX:+UseStringDeduplication" \
    "-XX:+UseSerialGC" \
    "-Xss512k" \
    "-XX:+ExitOnOutOfMemoryError" \
    "-jar" \
    "/app.jar"
