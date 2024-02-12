package com.egoodhall.nrpc.gen.generators;

import com.egoodhall.nrpc.gen.models.MethodInfo;
import com.egoodhall.nrpc.gen.models.ServiceInfo;
import com.egoodhall.nrpc.gen.util.TypeUtil;
import com.google.inject.Inject;
import com.squareup.javapoet.JavaFile;
import com.squareup.javapoet.MethodSpec;
import com.squareup.javapoet.ParameterSpec;
import com.squareup.javapoet.TypeSpec;
import java.util.List;
import java.util.Optional;
import javax.lang.model.element.Modifier;

public class ServiceGenerator extends FileGenerator {

  @Inject
  public ServiceGenerator(List<ServiceInfo> serviceInfos) {
    super(serviceInfos);
  }

  @Override
  protected Optional<JavaFile> generate(ServiceInfo serviceInfo) {
    TypeSpec.Builder builder = TypeSpec.interfaceBuilder(serviceInfo.getType());

    for (MethodInfo methodInfo : serviceInfo.getMethods()) {
      builder.addMethod(
        MethodSpec
          .methodBuilder(methodInfo.getName())
          .addModifiers(Modifier.PUBLIC, Modifier.ABSTRACT)
          .returns(TypeUtil.getResult(methodInfo.getReturnType()))
          .addParameter(
            ParameterSpec.builder(methodInfo.getParameterType(), "request").build()
          )
          .build()
      );
    }

    return Optional.of(
      JavaFile.builder(serviceInfo.getType().packageName(), builder.build()).build()
    );
  }
}
