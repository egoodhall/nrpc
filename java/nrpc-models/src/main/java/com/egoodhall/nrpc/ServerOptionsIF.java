package com.egoodhall.nrpc;

import java.util.Optional;
import org.immutables.value.Value;

@Value.Immutable
@NrpcImmutableStyle
public abstract class ServerOptionsIF implements Options {

  public static ServerOptions defaults() {
    return ServerOptions.builder().build();
  }

  @Value.Default
  public int getBufferSize() {
    return 256;
  }

  @Value.Default
  public ErrorHandler getErrorHandler() {
    return ignored -> {};
  }

  public abstract Optional<String> getQueueGroup();

  public abstract Optional<String> getNamespace();

  interface Builder extends Options.Builder<ServerOptions.Builder> {}
}
