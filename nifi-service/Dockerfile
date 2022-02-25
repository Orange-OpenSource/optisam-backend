FROM dockerproxy-iva.si.francetelecom.fr/apache/nifi:1.12.1
USER root
RUN chgrp -R 0 /opt/nifi/ \
  && chmod -R g+rwX /opt/nifi/

#COPY flow.xml.gz /opt/nifi/nifi-current/conf/flow.xml.gz  

#COPY start.sh /opt/nifi/nifi-current/conf/start.sh


#ENTRYPOINT ["/opt/nifi/nifi-current/conf/start.sh"]


ENTRYPOINT ["/nifi/scripts/start.sh"]
