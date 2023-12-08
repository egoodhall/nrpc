package com.egoodhall.nrpc;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import org.immutables.value.Value;

@Value.Style(
  get = {"get*", "is*", "are*"},
  init = "set*",
  clearBuilder = true,
  stagedBuilder = true,
  typeAbstract = "*IF",
  typeImmutable = "*",
  jacksonIntegration = false,
  deepImmutablesDetection = true,
  optionalAcceptNullable = true
)
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.SOURCE)
public @interface OptionsStyle {
}
