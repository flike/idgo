package server

func (s *Server) handleGet(r *Request) Reply {
	var idgen *MySQLIdGenerator
	var ok, isExist bool
	var id int64
	var err error

	if r.HasArgument(0) == false {
		return ErrNotEnoughArgs
	}

	idGenKey := string(r.Arguments[0])
	if len(idGenKey) == 0 {
		return ErrNoKey
	}
	s.Lock()
	idgen, ok = s.keyGeneratorMap[idGenKey]
	s.Unlock()

	if ok == false {
		isExist, err = s.IsKeyExist(idGenKey)
		if err != nil {
			return &ErrorReply{
				message: err.Error(),
			}
		}
		if isExist {
			idgen, err = NewMySQLIdGenerator(s.db, idGenKey)
			if err != nil {
				return &ErrorReply{
					message: err.Error(),
				}
			}
			s.Lock()
			s.keyGeneratorMap[idGenKey] = idgen
			s.Unlock()
		} else {
			return &IntReply{
				number: 0,
			}
		}
	}

	id, err = idgen.Next()
	if err != nil {
		return &ErrorReply{
			message: err.Error(),
		}
	}

	return &IntReply{
		number: id,
	}
}

//redis command(set abc 12)
func (s *Server) handleSet(r *Request) Reply {
	var idgen *MySQLIdGenerator
	var ok bool
	var err error

	if r.HasArgument(0) == false {
		return ErrNotEnoughArgs
	}

	idGenKey := string(r.Arguments[0])
	if len(idGenKey) == 0 {
		return ErrNoKey
	}
	idValue, errReply := r.GetInt(1)
	if errReply != nil {
		return errReply
	}
	s.Lock()
	idgen, ok = s.keyGeneratorMap[idGenKey]
	s.Unlock()
	if ok == false {
		idgen, err = NewMySQLIdGenerator(s.db, idGenKey)
		if err != nil {
			return &ErrorReply{
				message: err.Error(),
			}
		}
		s.Lock()
		s.keyGeneratorMap[idGenKey] = idgen
		s.Unlock()
	}

	err = idgen.Reset(idValue, false)
	if err != nil {
		return &ErrorReply{
			message: err.Error(),
		}
	}

	return &StatusReply{
		code: "OK",
	}
}

func (s *Server) handleExists(r *Request) Reply {
	var ok bool
	var id int64

	if r.HasArgument(0) == false {
		return ErrNotEnoughArgs
	}

	idGenKey := string(r.Arguments[0])
	if len(idGenKey) == 0 {
		return ErrNoKey
	}
	s.Lock()
	_, ok = s.keyGeneratorMap[idGenKey]
	s.Unlock()
	if ok {
		id = 1
	}

	return &IntReply{
		number: id,
	}
}

func (s *Server) handleDel(r *Request) Reply {
	var idgen *MySQLIdGenerator
	var ok bool

	if r.HasArgument(0) == false {
		return ErrNotEnoughArgs
	}

	idGenKey := string(r.Arguments[0])
	if len(idGenKey) == 0 {
		return ErrNoKey
	}
	s.Lock()
	idgen, ok = s.keyGeneratorMap[idGenKey]
	if ok {
		delete(s.keyGeneratorMap, idGenKey)
	}
	s.Unlock()
	if ok {
		err := idgen.DelKeyTable(idGenKey)
		if err != nil {
			return &ErrorReply{
				message: err.Error(),
			}
		}
	}

	return &StatusReply{
		code: "OK",
	}
}
