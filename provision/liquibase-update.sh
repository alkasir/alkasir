echo liquibase update
cd ${HOME}/src/github.com/alkasir/alkasir/central-db/

docker-compose up -d db

n=0
until [ $n -ge 20 ]
do
  PGPASSWORD=alkasir_central \
            psql \
            --host=127.0.0.1 \
            --port=39558 \
            --user=alkasir_central alkasir_central \
            -c "select 1"  \
    && break
  n=$[$n+1]
  echo retrying in 1 seconds...
  sleep 1
done

./liquibase update
