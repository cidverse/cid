############################################################
# Base Image
############################################################

# Base Image
FROM amazoncorretto:15-alpine

############################################################
# Installation
############################################################

# Copy files from rootfs to the container (there should only be one in /dist)
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
  "-Duser.language=${JVM_USER_LANGUAGE:-en}" \
  "-Duser.country=${JVM_USER_COUNTRY:-US}" \
  "-Duser.timezone=${JVM_USER_TIMEZONE:-UTC}" \
  "-Dorg.jboss.logging.provider=log4j2" \
  "-Dfile.encoding=${JVM_FILE_ENCODING:-UTF8}" \
  "${JAVA_OPTS_CUSTOM:--Dhello=world}" \
  "-XX:-TieredCompilation" \
  "-XX:+UseStringDeduplication" \
  "-XX:+UseSerialGC" \
  "-Xss512k" \
  "-XX:+ExitOnOutOfMemoryError" \
  "-jar" \
  "/app.jar"
