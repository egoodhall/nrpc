package com.egoodhall.nrpc;

import java.time.Duration;
import org.immutables.value.Value;

@Value.Immutable
@NrpcImmutableStyle
public abstract class ClientOptionsIF implements Options {

  public static ClientOptions defaults() {
    return ClientOptions.builder().build();
  }

  @Value.Default
  public Duration getTimeout() {
    return Duration.ofSeconds(10);
  }

  interface Builder extends Options.Builder<ClientOptions.Builder> {}
}
