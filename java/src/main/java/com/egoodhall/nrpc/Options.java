package com.egoodhall.nrpc;

import com.google.common.hash.Hashing;
import com.google.common.io.BaseEncoding;
import java.nio.charset.StandardCharsets;
import java.util.Optional;

public interface Options {
  Optional<String> getNamespace();

  interface Builder<T extends Builder<T>> {
    T setNamespace(String value);
    default T setHashNamespace(String... value) {
      return setNamespace(
        BaseEncoding
          .base16()
          .encode(
            Hashing
              .sha256()
              .hashString(String.join("", value), StandardCharsets.UTF_8)
              .asBytes()
          )
      );
    }
  }
}
