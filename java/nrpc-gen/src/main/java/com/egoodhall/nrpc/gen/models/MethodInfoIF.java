package com.egoodhall.nrpc.gen.models;

import com.egoodhall.nrpc.NrpcImmutableStyle;
import com.squareup.javapoet.TypeName;
import org.immutables.value.Value;

@Value.Immutable
@NrpcImmutableStyle
public interface MethodInfoIF {
  String getName();
  TypeName getReturnType();
  TypeName getParameterType();
}
