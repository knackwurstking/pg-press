package press

//func (h *Handler) HTMXGetCycleSummaryPDF(c echo.Context) error {
//	press, merr := h.getPressNumberFromParam(c)
//	if merr != nil {
//		return merr.Echo()
//	}
//
//	// Get cycle summary data using service
//	cycles, toolsMap, usersMap, merr := h.registry.PressCycles.GetCycleSummaryData(press)
//	if merr != nil {
//		return merr.Echo()
//	}
//
//	// Generate PDF
//	pdfBuffer, err := pdf.GenerateCycleSummaryPDF(press, cycles, toolsMap, usersMap)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "generate PDF"))
//	}
//
//	// Set response headers
//	filename := fmt.Sprintf("press_%d_cycle_summary_%s.pdf", press, time.Now().Format("2006-01-02"))
//	c.Response().Header().Set("Content-Type", "application/pdf")
//	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
//	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", pdfBuffer.Len()))
//
//	err = c.Stream(http.StatusOK, "application/pdf", pdfBuffer)
//	if err != nil {
//		return echo.NewHTTPError(http.StatusInternalServerError, errors.Wrap(err, "stream"))
//	}
//	return nil
//}
