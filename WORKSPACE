workspace(name = "com_github_prysmaticlabs_go_ssz")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "io_bazel_rules_go",
    urls = ["https://github.com/bazelbuild/rules_go/releases/download/0.18.6/rules_go-0.18.6.tar.gz"],
    sha256 = "f04d2373bcaf8aa09bccb08a98a57e721306c8f6043a2a0ee610fd6853dcde3d",
)

load("@io_bazel_rules_go//go:deps.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains(nogo = "@com_github_prysmaticlabs_go_ssz//:nogo")

http_archive(
    name = "bazel_gazelle",
    urls = ["https://github.com/bazelbuild/bazel-gazelle/releases/download/0.17.0/bazel-gazelle-0.17.0.tar.gz"],
    sha256 = "3c681998538231a2d24d0c07ed5a7658cb72bfb5fd4bf9911157c0e9ac6a2687",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

load("@com_github_prysmaticlabs_go_ssz//:deps.bzl", "go_ssz_dependencies")

go_ssz_dependencies()

http_archive(                                                                   
    name = "eth2_spec_tests",                                                   
    build_file_content = """                                                    
filegroup(                                                                      
    name = "test_data",                                                         
    srcs = glob([                                                               
        "**/*.yaml",                                                            
    ]),                                                                         
    visibility = ["//visibility:public"],                                       
)                                                                               
    """,                                                                        
    sha256 = "a531804ac35d2398d37cfa755a686280d8cb3a9649e993e3cf89640f06191d5e",
    url = "https://github.com/prysmaticlabs/eth2.0-spec-tests/releases/download/v0.8.1/base64_encoded_archive.tar.gz",
)   

http_archive(
    name = "io_kubernetes_build",
    sha256 = "dd02a62c2a458295f561e280411b04d2efbd97e4954986a401a9a1334cc32cc3",
    strip_prefix = "repo-infra-1b2ddaf3fb8775a5d0f4e28085cf846f915977a8",
    url = "https://github.com/kubernetes/repo-infra/archive/1b2ddaf3fb8775a5d0f4e28085cf846f915977a8.tar.gz",
)

go_repository(
    name = "com_github_golang_lint",
    commit = "5b3e6a55c961c61f4836ae6868c17b070744c590",
    importpath = "github.com/golang/lint",
)

go_repository(
    name = "org_golang_x_lint",
    commit = "5b3e6a55c961c61f4836ae6868c17b070744c590",
    importpath = "golang.org/x/lint",
)

go_repository(
    name = "com_github_ghodss_yaml",
    commit = "0ca9ea5df5451ffdf184b4428c902747c2c11cd7",  # v1.0.0
    importpath = "github.com/ghodss/yaml",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "51d6538a90f86fe93ac480b35f37b2be17fef232",  # v2.2.2
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "in_gopkg_d4l3k_messagediff_v1",
    commit = "29f32d820d112dbd66e58492a6ffb7cc3106312b",  # v1.2.1
    importpath = "gopkg.in/d4l3k/messagediff.v1",
)

# Do not add go dependencies here. They must be added in deps.bzl to provide
# dependencies to downstream bazel projects. Only test related dependencies
# are allowed in this WORKSPACE.
