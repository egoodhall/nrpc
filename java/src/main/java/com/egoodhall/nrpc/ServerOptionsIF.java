package com.egoodhall.nrpc;

import java.util.Optional;
import org.immutables.value.Value;

@Value.Immutable
@OptionsStyle
public interface ServerOptionsIF extends Options {
  default int getBufferSize() {
    return 256;
  }
  default ErrorHandler getErrorHandler() {
    return (ignored) -> {};
  }
  Optional<String> getQueueGroup();
  Optional<String> getNamespace();

  interface Builder extends Options.Builder<Builder> {}
}
