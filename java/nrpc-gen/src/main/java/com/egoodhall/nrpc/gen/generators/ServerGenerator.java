package com.egoodhall.nrpc.gen.generators;

import com.egoodhall.nrpc.ServerOptions;
import com.egoodhall.nrpc.gen.models.MethodInfo;
import com.egoodhall.nrpc.gen.models.ServiceInfo;
import com.google.common.base.CaseFormat;
import com.google.inject.Inject;
import com.google.protobuf.InvalidProtocolBufferException;
import com.squareup.javapoet.ClassName;
import com.squareup.javapoet.CodeBlock;
import com.squareup.javapoet.FieldSpec;
import com.squareup.javapoet.JavaFile;
import com.squareup.javapoet.MethodSpec;
import com.squareup.javapoet.ParameterSpec;
import com.squareup.javapoet.TypeName;
import com.squareup.javapoet.TypeSpec;
import io.nats.client.Connection;
import io.nats.service.Group;
import io.nats.service.Service;
import io.nats.service.ServiceEndpoint;
import io.nats.service.ServiceMessage;
import java.io.Closeable;
import java.util.List;
import java.util.Optional;
import javax.lang.model.element.Modifier;

public class ServerGenerator extends FileGenerator {

  private static final String CONNECTION_FIELD_NAME = "connection";
  private static final String OPTIONS_FIELD_NAME = "options";
  private static final String SERVICE_FIELD_NAME = "service";
  private static final String SERVER_FIELD_NAME = "server";

  @Inject
  public ServerGenerator(List<ServiceInfo> serviceInfos) {
    super(serviceInfos);
  }

  @Override
  protected Optional<JavaFile> generate(ServiceInfo serviceInfo) {
    ClassName serverType = ClassName.get(
      serviceInfo.getType().packageName(),
      serviceInfo.getType().simpleName() + "Server"
    );

    TypeSpec.Builder builder = TypeSpec
      .classBuilder(serverType)
      .addModifiers(Modifier.PUBLIC, Modifier.FINAL)
      .addSuperinterface(Closeable.class)
      .addFields(getFields(serviceInfo))
      .addMethod(getConstructor(serviceInfo))
      .addMethod(getCloseMethod());

    for (MethodInfo methodInfo : serviceInfo.getMethods()) {
      builder.addMethod(getServiceWrapperMethod(methodInfo));
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
        .builder(serviceInfo.getType(), SERVICE_FIELD_NAME)
        .addModifiers(Modifier.PRIVATE, Modifier.FINAL)
        .build(),
      FieldSpec
        .builder(ServerOptions.class, OPTIONS_FIELD_NAME)
        .addModifiers(Modifier.PRIVATE, Modifier.FINAL)
        .build(),
      FieldSpec
        .builder(Service.class, SERVER_FIELD_NAME)
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
        ParameterSpec.builder(serviceInfo.getType(), SERVICE_FIELD_NAME).build()
      )
      .addParameter(
        ParameterSpec.builder(ServerOptions.class, OPTIONS_FIELD_NAME).build()
      )
      .addStatement("this.$N = $N", CONNECTION_FIELD_NAME, CONNECTION_FIELD_NAME)
      .addStatement("this.$N = $N", SERVICE_FIELD_NAME, SERVICE_FIELD_NAME)
      .addStatement("this.$N = $N", OPTIONS_FIELD_NAME, OPTIONS_FIELD_NAME)
      .addStatement(
        "$T group = new $T($S)",
        ClassName.get(Group.class),
        ClassName.get(Group.class),
        serviceInfo.getType().simpleName()
      )
      .addStatement(
        "this.$N = $L",
        SERVER_FIELD_NAME,
        getNatsServiceCodeBlock(serviceInfo)
      )
      .addStatement("this.$N.startService()", SERVER_FIELD_NAME)
      .build();
  }

  private CodeBlock getNatsServiceCodeBlock(ServiceInfo serviceInfo) {
    String serviceBuilderName = "builder";
    CodeBlock.Builder builder = CodeBlock
      .builder()
      .add(
        CodeBlock
          .builder()
          .add("$T\n.builder()", ClassName.get(Service.class))
          .add("\n.name($S)", serviceInfo.getType().simpleName())
          .add("\n.version($S)", "0.0.0")
          .add("\n.connection($N)", CONNECTION_FIELD_NAME)
          .build()
      );

    for (MethodInfo methodInfo : serviceInfo.getMethods()) {
      builder
        .add("\n.addServiceEndpoint(\n")
        .add(getServiceEndpointCodeBlock(serviceInfo, methodInfo))
        .add("\n)");
    }

    return builder.add("\n.build()").build();
  }

  private CodeBlock getServiceEndpointCodeBlock(
    ServiceInfo serviceInfo,
    MethodInfo methodInfo
  ) {
    return CodeBlock
      .builder()
      .add("$T\n.builder()", ClassName.get(ServiceEndpoint.class))
      .add("\n.endpointName($S)", methodInfo.getName())
      .add(
        "\n.endpointSubject($N.getSubject($S, $S))",
        OPTIONS_FIELD_NAME,
        serviceInfo.getType().simpleName(),
        methodInfo.getName()
      )
      .add("\n.handler(this::$N)", getHandlerMethodName(methodInfo))
      .add("\n.build()")
      .build();
  }

  private MethodSpec getCloseMethod() {
    return MethodSpec
      .methodBuilder("close")
      .addModifiers(Modifier.PUBLIC)
      .returns(TypeName.VOID)
      .addAnnotation(Override.class)
      .addStatement("$N.stop()", SERVER_FIELD_NAME)
      .build();
  }

  private MethodSpec getServiceWrapperMethod(MethodInfo methodInfo) {
    String messageName = "message";

    return MethodSpec
      .methodBuilder(getHandlerMethodName(methodInfo))
      .addModifiers(Modifier.PRIVATE)
      .returns(TypeName.VOID)
      .addParameter(
        ParameterSpec.builder(ClassName.get(ServiceMessage.class), messageName).build()
      )
      .beginControlFlow("try")
      .addStatement(
        CodeBlock
          .builder()
          .add("$N.$N(", SERVICE_FIELD_NAME, methodInfo.getName())
          .add("$T.parseFrom($N.getData())", methodInfo.getParameterType(), messageName)
          .add(").consume(")
          .add("\nok -> message.respond($N, ok.toByteArray()),", CONNECTION_FIELD_NAME)
          .add(
            "\nerr -> message.respondStandardError($N, err.getMessage(), 500)",
            CONNECTION_FIELD_NAME
          )
          .add("\n)")
          .build()
      )
      .nextControlFlow("catch ($T e)", InvalidProtocolBufferException.class)
      .addStatement(
        "$N.respondStandardError($N, e.getMessage(), 400)",
        messageName,
        CONNECTION_FIELD_NAME
      )
      .nextControlFlow("catch ($T e)", Throwable.class)
      .addStatement(
        "$N.respondStandardError(connection, e.getMessage(), 500)",
        messageName
      )
      .addStatement("$N.getErrorHandler().handle(e)", OPTIONS_FIELD_NAME)
      .endControlFlow()
      .build();
  }

  private String getHandlerMethodName(MethodInfo methodInfo) {
    return (
      "handle" + CaseFormat.LOWER_CAMEL.to(CaseFormat.UPPER_CAMEL, methodInfo.getName())
    );
  }
}
