package com.egoodhall.nrpc.gen.util;

import com.egoodhall.nrpc.NrpcError;
import com.egoodhall.nrpc.Result;
import com.squareup.javapoet.ClassName;
import com.squareup.javapoet.ParameterizedTypeName;
import com.squareup.javapoet.TypeName;

public final class TypeUtil {

  public static final ClassName NRPC_ERROR = ClassName.get(NrpcError.class);
  private static final ClassName RESULT = ClassName.get(Result.class);

  private TypeUtil() {}

  public static ParameterizedTypeName getResult(TypeName ok, TypeName err) {
    return ParameterizedTypeName.get(RESULT, ok, err);
  }

  public static ParameterizedTypeName getResult(TypeName ok) {
    return getResult(ok, NRPC_ERROR);
  }
}
