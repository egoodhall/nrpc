package com.egoodhall.nrpc.gen.generators;

import com.egoodhall.nrpc.ClientOptions;
import com.egoodhall.nrpc.Results;
import com.egoodhall.nrpc.gen.models.MethodInfo;
import com.egoodhall.nrpc.gen.models.ServiceInfo;
import com.egoodhall.nrpc.gen.util.TypeUtil;
import com.google.common.base.CaseFormat;
import com.google.inject.Inject;
import com.squareup.javapoet.ClassName;
import com.squareup.javapoet.CodeBlock;
import com.squareup.javapoet.FieldSpec;
import com.squareup.javapoet.JavaFile;
import com.squareup.javapoet.MethodSpec;
import com.squareup.javapoet.ParameterSpec;
import com.squareup.javapoet.TypeSpec;
import io.nats.client.Connection;
import java.util.List;
import java.util.Optional;
import javax.lang.model.element.Modifier;

public class ClientGenerator extends FileGenerator {

  private static final String CONNECTION_FIELD_NAME = "connection";
  private static final String OPTIONS_FIELD_NAME = "options";
  private static final String SERVICE_FIELD_NAME = "service";
  private static final String SERVER_FIELD_NAME = "server";

  @Inject
  public ClientGenerator(List<ServiceInfo> serviceInfos) {
    super(serviceInfos);
  }

  @Override
  protected Optional<JavaFile> generate(ServiceInfo serviceInfo) {
    ClassName serverType = ClassName.get(
      serviceInfo.getType().packageName(),
      serviceInfo.getType().simpleName() + "Client"
    );

    TypeSpec.Builder builder = TypeSpec
      .classBuilder(serverType)
      .addModifiers(Modifier.PUBLIC, Modifier.FINAL)
      .addFields(getFields(serviceInfo))
      .addMethod(getConstructor(serviceInfo));

    for (MethodInfo methodInfo : serviceInfo.getMethods()) {
      builder.addMethod(getClientMethod(serviceInfo, methodInfo));
    }

    return Optional.of(
      JavaFile.builder(serverType.packageName(), builder.build()).build()
    );
  }

  private List<FieldSpec> getFields(ServiceInfo serviceInfo) {
    return List.of(
      FieldSpec
        .builder(Connection.class, CONNECTION_FIELD_NAME)
        .addModifiers(Modifier.PRIVATE, Modifier.FINAL)
        .build(),
      FieldSpec
        .builder(ClientOptions.class, OPTIONS_FIELD_NAME)
        .addModifiers(Modifier.PRIVATE, Modifier.FINAL)
        .build()
    );
  }

  private MethodSpec getConstructor(ServiceInfo serviceInfo) {
    return MethodSpec
      .constructorBuilder()
      .addModifiers(Modifier.PUBLIC)
      .addParameter(
        ParameterSpec.builder(Connection.class, CONNECTION_FIELD_NAME).build()
      )
      .addParameter(
        ParameterSpec.builder(ClientOptions.class, OPTIONS_FIELD_NAME).build()
      )
      .addStatement("this.$N = $N", CONNECTION_FIELD_NAME, CONNECTION_FIELD_NAME)
      .addStatement("this.$N = $N", OPTIONS_FIELD_NAME, OPTIONS_FIELD_NAME)
      .build();
  }

  private MethodSpec getClientMethod(ServiceInfo serviceInfo, MethodInfo methodInfo) {
    String requestName = "request";

    return MethodSpec
      .methodBuilder(methodInfo.getName())
      .addModifiers(Modifier.PUBLIC)
      .returns(TypeUtil.getResult(methodInfo.getReturnType(), TypeUtil.NRPC_ERROR))
      .addParameter(
        ParameterSpec.builder(methodInfo.getParameterType(), requestName).build()
      )
      .addStatement(
        CodeBlock
          .builder()
          .add("return $N.requestWithTimeout(", CONNECTION_FIELD_NAME)
          .add(
            "\n$N.getSubject($S, $S),",
            OPTIONS_FIELD_NAME,
            serviceInfo.getType().simpleName(),
            methodInfo.getName()
          )
          .add("\n$N.toByteArray(),", requestName)
          .add("\n$N.getTimeout()", OPTIONS_FIELD_NAME)
          .add("\n)")
          .add(
            "\n.thenApply(m -> $T.fromMessage(m, $T::parseFrom))",
            Results.class,
            methodInfo.getReturnType()
          )
          .add("\n.exceptionally($T::fromThrowable)", Results.class)
          .add("\n.join()")
          .build()
      )
      .build();
  }

  private String getHandlerMethodName(MethodInfo methodInfo) {
    return (
      "handle" + CaseFormat.LOWER_CAMEL.to(CaseFormat.UPPER_CAMEL, methodInfo.getName())
    );
  }
}
