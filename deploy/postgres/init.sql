-- Este script solo corre en el primer arranque del contenedor, con el
-- volumen vacío. Si lo cambias más tarde, hace falta `make clean` para
-- que Postgres lo vuelva a ejecutar.

-- catalog
CREATE USER catalog WITH PASSWORD 'catalog';
CREATE DATABASE catalog OWNER catalog;
REVOKE CONNECT ON DATABASE catalog FROM PUBLIC; -- solo su propio dueño se conecta

-- betting
CREATE USER betting WITH PASSWORD 'betting';
CREATE DATABASE betting OWNER betting;
REVOKE CONNECT ON DATABASE betting FROM PUBLIC;

-- treasury
CREATE USER treasury WITH PASSWORD 'treasury';
CREATE DATABASE treasury OWNER treasury;
REVOKE CONNECT ON DATABASE treasury FROM PUBLIC;

-- payout
CREATE USER payout WITH PASSWORD 'payout';
CREATE DATABASE payout OWNER payout;
REVOKE CONNECT ON DATABASE payout FROM PUBLIC;
