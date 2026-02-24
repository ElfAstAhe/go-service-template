#!/usr/bin/env bash

cd ..

mkdir -p api/{grpc,proto/example-service,rest} \
    cmd/example-service \
    configs \
    deployments \
    docs \
    internal/{app,config,domain,usecase,repository/postgres,transport/rest,transport/grpc} \
    migrations/example-service \
    pkg/{auth,config,db,domain,errs,helper,logger,migration/goose,repository,transport/middleware,utils} \
    scripts

cd scripts
