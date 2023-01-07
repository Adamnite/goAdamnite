# Consensus


## Abstract
The Adamnite Consensus Protocol is an implementation of the Delegated Proof of Stake consensus algorithm, designed to allow for the current state of the Adamnite Protocol to come to agreement on the current state of the ledger in an efficient and secure manner. 

## Description
We designate a participant as a node that has a set of identities, represented by addresses. Each address has a predefined balance of NITE, the native currency of the Adamnite Blockchain. NITE is used to participate in consensus: a participant participates in consensus by either convincing other participants to vote for them, or by voting for other participants. We define a "witness_candidate" as a particpant who has had at least one other participant stake any NITE to them. A "witness", on the other hand, is a "witness_candiate" who was actually selected to propose or approve blocks for the current round. A "round" is defined as an epoch lasting 162 blocks: each round, a new set of witnesses are selected. The reputation of a witness_candiate is a score that summarizes their past behavior. It is comprised of three weights: the amount of times they have been selected as a witness, their history of proposing or approving blocks that are actually approved and finalized, and their consistency/availablity. If a witness is found to be proposing blocks on two different forks of the current ledger or proposes a nonsense/spam block, their reputation score decreases significantly and they are replaced with another candidate. A more formal description can be found in the technical paper.


## File Recommendations
Files that define the general DPOS process, the start of a new round, a witness pool for organizing all of the witness candidates, and the actual agreement process should be defined. Self-selection using VRFs (more information in the Protocol Implementation Document and Technical Paper) should also be defined as a seperate file; in short, witness_candidates individually calculate a score using the VRF function and then check to see if they have been selected for consensus at the beginning of every round. Reputations and fork choice rule (defined in the technical paper) should also be implemented. 

Again, the focus is on orignality and then the implementation of additional features. 



