package helpers

import (
	"github.com/prysmaticlabs/go-ssz"
	p2ppb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	pb "github.com/prysmaticlabs/prysm/proto/beacon/p2p/v1"
	"github.com/prysmaticlabs/prysm/shared/params"
)

// ForkVersionByteLength length of fork version byte array.
const ForkVersionByteLength = 4

// DomainByteLength length of domain byte array.
const DomainByteLength = 4

// ComputeSigningRoot computes the root of the object by calculating the root of the object domain tree.
//
// Spec pseudocode definition:
//	def compute_signing_root(ssz_object: SSZObject, domain: Domain) -> Root:
//    """
//    Return the signing root of an object by calculating the root of the object-domain tree.
//    """
//    domain_wrapped_object = SigningRoot(
//        object_root=hash_tree_root(ssz_object),
//        domain=domain,
//    )
//    return hash_tree_root(domain_wrapped_object)
func ComputeSigningRoot(object interface{}, domain []byte) ([32]byte, error) {
	objRoot, err := ssz.HashTreeRoot(object)
	if err != nil {
		return [32]byte{}, err
	}
	container := &p2ppb.SigningRoot{
		ObjectRoot: objRoot[:],
		Domain:     domain,
	}
	return ssz.HashTreeRoot(container)
}

// ComputeDomain returns the domain version for BLS private key to sign and verify with a zeroed 4-byte
// array as the fork version.
//
// def compute_domain(domain_type: DomainType, fork_version: Version=None, genesis_validators_root: Root=None) -> Domain:
//    """
//    Return the domain for the ``domain_type`` and ``fork_version``.
//    """
//    if fork_version is None:
//        fork_version = GENESIS_FORK_VERSION
//    if genesis_validators_root is None:
//        genesis_validators_root = Root()  # all bytes zero by default
//    fork_data_root = compute_fork_data_root(fork_version, genesis_validators_root)
//    return Domain(domain_type + fork_data_root[:28])
func ComputeDomain(domainType [DomainByteLength]byte, forkVersion []byte, genesisValidatorsRoot []byte) ([]byte, error) {
	if forkVersion == nil {
		forkVersion = params.BeaconConfig().GenesisForkVersion
	}
	if genesisValidatorsRoot == nil {
		genesisValidatorsRoot = params.BeaconConfig().ZeroHash[:]
	}
	forkBytes := [ForkVersionByteLength]byte{}
	copy(forkBytes[:], forkVersion)

	forkDataRoot, err := computeForkDataRoot(forkVersion, genesisValidatorsRoot)
	if err != nil {
		return nil, err
	}

	return domain(domainType, forkDataRoot[:]), nil
}

// this returns the bls domain given by the domain type and fork data root..
func domain(domainType [DomainByteLength]byte, forkDataRoot []byte) []byte {
	b := []byte{}
	b = append(b, domainType[:4]...)
	b = append(b, forkDataRoot[:28]...)
	return b
}

// this returns the 32byte fork data root for the ``current_version`` and ``genesis_validators_root``.
// This is used primarily in signature domains to avoid collisions across forks/chains.
//
// Spec pseudocode definition:
//	def compute_fork_data_root(current_version: Version, genesis_validators_root: Root) -> Root:
//    """
//    Return the 32-byte fork data root for the ``current_version`` and ``genesis_validators_root``.
//    This is used primarily in signature domains to avoid collisions across forks/chains.
//    """
//    return hash_tree_root(ForkData(
//        current_version=current_version,
//        genesis_validators_root=genesis_validators_root,
//    ))
func computeForkDataRoot(version []byte, root []byte) ([32]byte, error) {
	r, err := ssz.HashTreeRoot(&pb.ForkData{
		CurrentVersion:        version,
		GenesisValidatorsRoot: root,
	})
	if err != nil {
		return [32]byte{}, err
	}
	return r, nil
}
