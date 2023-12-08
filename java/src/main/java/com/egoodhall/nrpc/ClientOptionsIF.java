package com.egoodhall.nrpc;

import java.time.Duration;
import org.immutables.value.Value;

@Value.Immutable
@OptionsStyle
public interface ClientOptionsIF extends Options {
  @Value.Default
  default Duration getTimeout() {
    return Duration.ofSeconds(30);
  }

  interface Builder extends Options.Builder<Builder> {}
}
