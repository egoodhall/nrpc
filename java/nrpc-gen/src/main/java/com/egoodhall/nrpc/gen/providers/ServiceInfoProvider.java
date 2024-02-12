package com.egoodhall.nrpc.gen.providers;

import static com.google.protobuf.DescriptorProtos.FileDescriptorProto;
import static com.google.protobuf.DescriptorProtos.MethodDescriptorProto;
import static com.google.protobuf.DescriptorProtos.ServiceDescriptorProto;

import com.egoodhall.nrpc.gen.models.MethodInfo;
import com.egoodhall.nrpc.gen.models.ServiceInfo;
import com.google.common.collect.ImmutableList;
import com.google.inject.Inject;
import com.google.inject.Provider;
import com.google.inject.Singleton;
import com.google.protobuf.compiler.PluginProtos;
import com.salesforce.jprotoc.ProtoTypeMap;
import com.squareup.javapoet.ClassName;
import java.util.List;

@Singleton
public class ServiceInfoProvider implements Provider<List<ServiceInfo>> {

  private final PluginProtos.CodeGeneratorRequest request;
  private final ProtoTypeMap protoTypeMap;

  @Inject
  public ServiceInfoProvider(
    PluginProtos.CodeGeneratorRequest request,
    ProtoTypeMap protoTypeMap
  ) {
    this.request = request;
    this.protoTypeMap = protoTypeMap;
  }

  @Override
  public List<ServiceInfo> get() {
    ImmutableList.Builder<ServiceInfo> builder = ImmutableList.builder();
    for (FileDescriptorProto fileDescriptor : request.getProtoFileList()) {
      String javaPackage = fileDescriptor.getOptions().getJavaPackage();
      for (ServiceDescriptorProto serviceDescriptor : fileDescriptor.getServiceList()) {
        ServiceInfo.Builder serviceBuilder = ServiceInfo
          .builder()
          .setType(ClassName.get(javaPackage, serviceDescriptor.getName()));

        for (MethodDescriptorProto methodDescriptor : serviceDescriptor.getMethodList()) {
          serviceBuilder.addMethods(parseMethod(methodDescriptor));
        }

        builder.add(serviceBuilder.build());
      }
    }
    return builder.build();
  }

  private MethodInfo parseMethod(MethodDescriptorProto methodDescriptor) {
    return MethodInfo
      .builder()
      .setName(methodDescriptor.getName())
      .setParameterType(getType(methodDescriptor.getInputType()))
      .setReturnType(getType(methodDescriptor.getOutputType()))
      .build();
  }

  private ClassName getType(String protoTypeName) {
    String javaTypeName = protoTypeMap.toJavaTypeName(protoTypeName);
    int lastDotIdx = javaTypeName.lastIndexOf('.');
    return ClassName.get(
      javaTypeName.substring(0, lastDotIdx),
      javaTypeName.substring(lastDotIdx + 1)
    );
  }
}
