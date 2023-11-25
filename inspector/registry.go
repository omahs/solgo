package inspector

var registry = map[DetectorType]Detector{}

func DetectorExists(detectorType DetectorType) bool {
	_, ok := registry[detectorType]
	return ok
}

func GetDetector(detectorType DetectorType) Detector {
	return registry[detectorType]
}

func RegisterDetector(detectorType DetectorType, detector Detector) bool {
	if !DetectorExists(detectorType) {
		registry[detectorType] = detector
		return true
	}

	return false
}

func (i *Inspector) RegisterDetectors() {
	RegisterDetector(StateVariableDetectorType, NewStateVariableDetector(i.ctx, i))
	RegisterDetector(TransferDetectorType, NewTransferDetector(i.ctx, i))
	RegisterDetector(MintDetectorType, NewMintDetector(i.ctx, i))
	/*
		 	RegisterDetector(TransferDetector, &TransferDetectorImpl{})
			RegisterDetector(MintDetector, &MintDetectorImpl{})
			RegisterDetector(BurnDetector, &BurnDetectorImpl{})
	*/
}
