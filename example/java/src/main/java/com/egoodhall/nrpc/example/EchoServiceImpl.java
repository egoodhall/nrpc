package com.egoodhall.nrpc.example;

import com.egoodhall.nrpc.NrpcError;
import com.egoodhall.nrpc.Result;

public class EchoServiceImpl implements EchoService {

  public EchoServiceImpl() {}

  @Override
  public Result<EchoReply, NrpcError> echo(EchoRequest request) {
    return Result.ok(EchoReply.newBuilder().setMessage(request.getMessage()).build());
  }
}
