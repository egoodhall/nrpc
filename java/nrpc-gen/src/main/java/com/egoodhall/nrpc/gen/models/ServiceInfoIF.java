package com.egoodhall.nrpc.gen.models;

import com.egoodhall.nrpc.NrpcImmutableStyle;
import com.squareup.javapoet.ClassName;
import java.util.List;
import org.immutables.value.Value;

@Value.Immutable
@NrpcImmutableStyle
public interface ServiceInfoIF {
  ClassName getType();
  List<MethodInfo> getMethods();
}
