package com.egoodhall.nrpc.gen.generators;

import com.egoodhall.nrpc.gen.models.ServiceInfo;
import com.google.common.collect.ImmutableList;
import com.squareup.javapoet.JavaFile;
import java.util.List;
import java.util.Optional;

public abstract class FileGenerator {

  private final List<ServiceInfo> serviceInfos;

  public FileGenerator(List<ServiceInfo> serviceInfos) {
    this.serviceInfos = serviceInfos;
  }

  public final List<JavaFile> generate() {
    ImmutableList.Builder<JavaFile> javaFiles = ImmutableList.builder();
    for (ServiceInfo serviceInfo : serviceInfos) {
      generate(serviceInfo).ifPresent(javaFiles::add);
    }
    return javaFiles.build();
  }

  protected abstract Optional<JavaFile> generate(ServiceInfo serviceInfo);
}
