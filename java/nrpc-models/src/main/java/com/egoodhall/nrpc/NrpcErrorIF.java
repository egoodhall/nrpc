package com.egoodhall.nrpc;

import org.immutables.value.Value;

@Value.Immutable
@NrpcImmutableStyle
public interface NrpcErrorIF {
  @Value.Parameter(order = 0)
  int getCode();

  @Value.Parameter(order = 1)
  String getMessage();
}
