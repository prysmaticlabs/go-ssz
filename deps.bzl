"""
go-ssz dependencies.

Add go_repository dependencies where with the _maybe function so that downstream
bazel projects' go_repository take precidence over the commits specified here.

Add the license comment to each dependency for quick analysis of the licensing
requirements for this project.
"""
load("@bazel_gazelle//:deps.bzl", "go_repository")

def go_ssz_dependencies():
    _maybe(
        # MIT License
        go_repository,
        name = "com_github_karlseguin_ccache",
        commit = "ec06cd93a07565b373789b0078ba88fe697fddd9",
        importpath = "github.com/karlseguin/ccache",
    )

    _maybe(
        # Apache License 2.0
        go_repository,
        name = "com_github_prometheus_client_golang",
        commit = "662e8a9ffaaa74a4d050023c2cb26902cd9bab63",
        importpath = "github.com/prometheus/client_golang",
    )

    _maybe(
        # Apache License 2.0
        go_repository,
        name = "com_github_prometheus_common",
        commit = "1ba88736f028e37bc17328369e94a537ae9e0234",
        importpath = "github.com/prometheus/common",
    )

    _maybe(
        # Apache License 2.0
        go_repository,
        name = "com_github_prometheus_client_model",
        commit = "fd36f4220a901265f90734c3183c5f0c91daa0b8",
        importpath = "github.com/prometheus/client_model",
    )

    _maybe(
        # Apache License 2.0
        go_repository,
        name = "com_github_prometheus_procfs",
        commit = "bbced9601137e764853b2fad7ec3e2dc4c504e02",
        importpath = "github.com/prometheus/procfs",
    )

    _maybe(
        # Apache License 2.0
        go_repository,
        name = "com_github_matttproud_golang_protobuf_extensions",
        commit = "c12348ce28de40eed0136aa2b644d0ee0650e56c", 
        importpath = "github.com/matttproud/golang_protobuf_extensions",
    )

    _maybe(
        # MIT License
        go_repository,
        name = "com_github_beorn7_perks",
        commit = "4ded152d4a3e2847f17f185a27b2041ae7b63979",
        importpath = "github.com/beorn7/perks",
    )

    _maybe(
        # GNU Lesser General Public License v3.0
        go_repository,
        name = "com_github_ethereum_go_ethereum",
        commit = "099afb3fd89784f9e3e594b7c2ed11335ca02a9b",
        importpath = "github.com/ethereum/go-ethereum",
        # Note: go-ethereum is not bazel-friendly with regards to cgo. We have a
        # a fork that has resolved these issues by disabling HID/USB support and
        # some manual fixes for c imports in the crypto package. This is forked
        # branch should be updated from time to time with the latest go-ethereum
        # code.
        remote = "https://github.com/prysmaticlabs/bazel-go-ethereum",
        vcs = "git",
    )

    _maybe(
        # Permissive license
        # https://github.com/golang/crypto/blob/master/LICENSE
        go_repository,
        name = "org_golang_x_crypto",
        commit = "8dd112bcdc25174059e45e07517d9fc663123347",
        importpath = "golang.org/x/crypto",
    )

    _maybe(
        # Apache License 2.0
        # https://github.com/prysmaticlabs/go-bitfield/blob/master/LICENSE
        go_repository,
        name = "com_github_prysmaticlabs_go_bitfield",
        commit = "ec88cc4d1d143cad98308da54b73d0cdb04254eb",
        importpath = "github.com/prysmaticlabs/go-bitfield",
    )

    _maybe(
        go_repository,
        name = "com_github_minio_sha256_simd",
        commit = "05b4dd3047e5d6e86cb4e0477164b850cd896261",
        importpath = "github.com/minio/sha256-simd",
    )

    _maybe(
        go_repository,
        name = "com_github_protolambda_zssz",
        commit = "632f11e5e281660402bd0ac58f76090f3503def0",
        importpath = "github.com/protolambda/zssz",
    )

    _maybe(
        go_repository,
        name = "com_github_pkg_errors",
        commit = "27936f6d90f9c8e1145f11ed52ffffbfdb9e0af7",
        importpath = "github.com/pkg/errors",
    )

def _maybe(repo_rule, name, **kwargs):
    if name not in native.existing_rules():
        repo_rule(name = name, **kwargs)
