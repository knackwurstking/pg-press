# Admin Tools Section - Usage Examples

## Overview

This document provides practical examples of using the Admin Tools section to detect and resolve tool overlap issues across presses.

## Accessing the Admin Section

1. Navigate to the main tools page: `/tools`
2. At the top of the page, you'll see a section titled "Admin: Werkzeug-Überschneidungen"
3. Click on the section header to expand it
4. The system will automatically load and analyze tool overlaps across all presses

## Example Scenarios

### Scenario 1: No Overlaps Detected (Ideal State)

**Situation**: All tools are properly managed with no conflicts

**Display**:

```
✓ Keine Überschneidungen gefunden
Alle Werkzeuge sind korrekt auf einzelne Pressen beschränkt.
```

**Interpretation**: System is healthy, all tool assignments are correct.

### Scenario 2: Single Tool Overlap

**Situation**: Tool ID 1234 appears on both Press 2 and Press 4 simultaneously

**Display**:

```
⚠️ 1 Werkzeuge gefunden, die gleichzeitig auf mehreren Pressen verwendet werden.

[Error Card]
Format 400x300 Code ABC123
Tool ID: 1234 | Zeitraum: 15.03.2024 - 20.03.2024

Betroffene Pressen:

Presse 2                    Presse 4
Position: Oben             Position: Top Cassette
Start: 15.03.2024 10:30    Start: 18.03.2024 14:15
Ende: 25.03.2024 16:45     Ende: 20.03.2024 09:30
Dauer: 10 Tage, 6 Stunden  Dauer: 1 Tage, 19 Stunden
```

**Interpretation**:

- Tool 1234 was active on Press 2 from March 15-25
- Same tool was simultaneously active on Press 4 from March 18-20
- Physical impossibility: one tool cannot be in two places
- Likely causes: data entry error, missing tool change record

### Scenario 3: Multiple Tool Overlaps

**Situation**: Several tools have overlapping periods

**Display**:

```
⚠️ 3 Werkzeuge gefunden, die gleichzeitig auf mehreren Pressen verwendet werden.

[Error Card 1]
Format 500x400 Code DEF456
Tool ID: 2345 | Zeitraum: 01.04.2024 - 05.04.2024

[Error Card 2]
Format 300x200 Code GHI789
Tool ID: 3456 | Zeitraum: 08.04.2024 - 12.04.2024

[Error Card 3]
Format 600x350 Code JKL012
Tool ID: 4567 | Zeitraum: 15.04.2024 - 18.04.2024
```

**Interpretation**: Multiple data integrity issues requiring systematic investigation.

## Resolution Workflow

### Step 1: Identify the Problem

1. Open the Admin Tools section
2. Review all reported overlaps
3. Note the specific dates and presses involved
4. Document the tool IDs and codes affected

### Step 2: Investigate Root Cause

For each overlapping tool:

1. **Check Press Logs**:
   - Go to individual press pages (`/tools/press/2`, `/tools/press/4`, etc.)
   - Review "Pressennutzungsverlauf" (Press Usage History)
   - Look for cycle entries during the overlap period

2. **Verify with Operators**:
   - Contact operators working during the overlap period
   - Confirm actual tool changes and timing
   - Identify any manual tool swaps not recorded

3. **Review Data Entry**:
   - Check for backdated cycle entries
   - Look for bulk data imports during the period
   - Identify potential timing errors

### Step 3: Resolve Data Issues

**Scenario A: Incorrect End Date**

```
Problem: Tool removed from Press 2 on March 18, but cycles continued until March 25
Solution: Update the last cycle entry for Press 2 to March 18
```

**Scenario B: Missing Tool Change Record**

```
Problem: Tool moved from Press 2 to Press 4 on March 18, but no tool change recorded
Solution: Add tool change entry showing removal from Press 2 and installation on Press 4
```

**Scenario C: Duplicate Cycle Entries**

```
Problem: Same cycles entered twice with different dates
Solution: Remove duplicate entries, keep correct dates
```

### Step 4: Verify Resolution

1. After making corrections, refresh the Admin Tools section
2. Confirm overlaps are resolved
3. Check that cycle summaries show correct tool transitions
4. Document resolution in system notes

## Best Practices

### Regular Monitoring

- **Daily Check**: Quick review during morning startup
- **Weekly Audit**: Detailed analysis of any detected overlaps
- **Monthly Review**: Pattern analysis for recurring issues

### Prevention Strategies

- **Immediate Recording**: Enter tool changes as they happen
- **Double-Check Dates**: Verify dates when entering historical data
- **Operator Training**: Ensure all staff understand proper recording procedures

### Data Quality Maintenance

- **Consistent Procedures**: Standardize tool change documentation
- **Regular Backups**: Maintain data backups before bulk operations
- **Audit Trail**: Keep records of all data corrections made

## Troubleshooting Common Issues

### Admin Section Won't Load

**Symptoms**: Spinner continues indefinitely, no content appears
**Causes**: Database connectivity, permissions, server errors
**Solutions**:

1. Refresh the page
2. Check browser console for JavaScript errors
3. Contact system administrator if problem persists

### Unexpected Overlaps Detected

**Symptoms**: Overlaps shown for tools known to be correct
**Causes**: Data synchronization delays, caching issues
**Solutions**:

1. Wait 5-10 minutes and refresh
2. Check individual press pages for recent updates
3. Verify system clock synchronization

### Performance Issues

**Symptoms**: Slow loading, timeouts
**Causes**: Large datasets, database performance
**Solutions**:

1. Try during off-peak hours
2. Contact administrator about database optimization
3. Consider archiving old data

## Integration with Existing Workflows

### Tool Change Process

1. **Before**: Check Admin Tools for current tool status
2. **During**: Record tool change immediately in system
3. **After**: Verify no new overlaps appear in Admin Tools

### Maintenance Scheduling

1. Use Admin Tools to identify tools with frequent overlaps
2. Schedule maintenance during periods with fewer conflicts
3. Monitor tool performance correlation with data quality

### Quality Control

1. Include Admin Tools check in quality procedures
2. Address any overlaps before shift handoffs
3. Document resolution steps for future reference

## Advanced Usage Tips

### Interpreting Duration Information

- **Short Overlaps** (< 24 hours): Likely timing errors or normal same-day changes
- **Medium Overlaps** (1-7 days): Probably data entry mistakes
- **Long Overlaps** (> 1 week): Serious data integrity issues requiring investigation

### Pattern Recognition

- **Regular Overlaps**: May indicate systemic process issues
- **Specific Tools**: Some tools may be more prone to recording errors
- **Particular Presses**: Certain presses may need procedure improvements

### Historical Analysis

- **Trend Monitoring**: Track whether overlaps are increasing or decreasing
- **Problem Tools**: Identify tools that frequently show overlaps
- **Training Needs**: Correlate overlaps with operator training requirements
