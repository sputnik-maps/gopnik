	"container/list"
	"sync"
type saveQueueElem struct {
}



	if err != nil {
		log.Error("Cache write error: %v", err)
	}
}

	srv.saveListMu.Lock()
	defer srv.saveListMu.Unlock()

	elem := saveQueueElem{
		TileCoord: coord,
	}
	return srv.saveList.PushFront(&elem)
}

func (srv *TileServer) saveQueueRemove(elem *list.Element) {
	srv.saveListMu.Lock()
	defer srv.saveListMu.Unlock()

	srv.saveList.Remove(elem)
}

	srv.saveListMu.RLock()
	defer srv.saveListMu.RUnlock()

	for e := srv.saveList.Front(); e != nil; e = e.Next() {
		elem := e.Value.(*saveQueueElem)
			return elem.Data
		}
	}
	return nil
}


	data := srv.saveQueueGet(metacoord)
	if data == nil {
		return nil
	}

}

	if data := srv.checkSaveQueue(tc); data != nil {
		w.Header().Set("Content-Type", "image/png")
		_, err := w.Write(data)
		if err != nil {
			log.Warning("HTTP Write error: %v", err)
		}
		return
	}



