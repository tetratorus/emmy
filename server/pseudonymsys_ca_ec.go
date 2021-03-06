/*
 * Copyright 2017 XLAB d.o.o.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package server

import (
	"math/big"

	"github.com/xlab-si/emmy/config"
	"github.com/xlab-si/emmy/crypto/ecpseudsys"
	pb "github.com/xlab-si/emmy/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) GenerateCertificate_EC(stream pb.PseudonymSystemCA_GenerateCertificate_ECServer) error {
	req, err := s.receive(stream)
	if err != nil {
		return err
	}

	d := config.LoadPseudonymsysCASecret()
	pubKey := config.LoadPseudonymsysCAPubKey()
	ca := ecpseudsys.NewCA(d, pubKey, curve)

	sProofRandData := req.GetSchnorrEcProofRandomData()
	x := sProofRandData.X.GetNativeType()
	a := sProofRandData.A.GetNativeType()
	b := sProofRandData.B.GetNativeType()

	challenge := ca.GetChallenge(a, b, x)
	resp := &pb.Message{
		Content: &pb.Message_Bigint{
			&pb.BigInt{
				X1: challenge.Bytes(),
			},
		},
	}

	if err := s.send(resp, stream); err != nil {
		return err
	}

	req, err = s.receive(stream)
	if err != nil {
		return err
	}

	sProofData := req.GetSchnorrProofData()
	z := new(big.Int).SetBytes(sProofData.Z)
	cert, err := ca.Verify(z)

	if err != nil {
		s.Logger.Debug(err)
		return status.Error(codes.Internal, err.Error())
	}

	resp = &pb.Message{
		Content: &pb.Message_PseudonymsysCaCertificateEc{
			&pb.PseudonymsysCACertificateEC{
				BlindedA: pb.ToPbECGroupElement(cert.BlindedA),
				BlindedB: pb.ToPbECGroupElement(cert.BlindedB),
				R:        cert.R.Bytes(),
				S:        cert.S.Bytes(),
			},
		},
	}

	if err = s.send(resp, stream); err != nil {
		return err
	}

	return nil
}
